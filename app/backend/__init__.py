import aiohttp
import datetime
import flask
import json
import os

from flask import Flask, render_template, redirect, request, session
from flask_bootstrap import Bootstrap5
from flask_wtf import FlaskForm, CSRFProtect
from flask_session import Session
from flask_sqlalchemy import SQLAlchemy
from wtforms import StringField, SubmitField

def load_credentials(file_path):
    if os.path.isfile(file_path):
        with open(file_path, 'r') as secret:
            data = json.load(secret)
            return {
                "GOOGLE_CONFIG":  { k.upper(): v for k, v in data["installed"].items() }
            }
    else:
        return {}


def create_app(app_endpoint='http://localhost', port=5000):
    app = Flask(__name__, instance_relative_config=True)
    app.config.from_mapping(
        SECRET_KEY='dev',
        API_BASE='http://localhost:3000',
        PORT=port,
        GOOGLE_CONFIG_FILE="./client_secret.json",
        GOOGLE_REDIRECT_URI=f'{app_endpoint}:{port}',
        SQLALCHEMY_DATABASE_URI='postgresql://postgres:db712bccc14d602212c928a39ba7e23d@localhost:5432/yourtube_app',
        SESSION_TYPE='sqlalchemy',
    )

    db = SQLAlchemy(app)

    app.config.from_mapping(SESSION_SQLALCHEMY=db)
    app.config.from_prefixed_env(prefix='YOURTUBE')
    app.config.from_mapping(**load_credentials(app.config['GOOGLE_CONFIG_FILE']))

    bootstrap = Bootstrap5(app)
    csrf = CSRFProtect(app)
    Session(app)

    return app, db, bootstrap, csrf

app, db, booetstrap, csrf = create_app()

def published_since(days=2):
    since = datetime.datetime.today() - datetime.timedelta(days=2)

    return since.strftime("%Y-%m-%d")  

class FilterForm(FlaskForm):
    min_duration = StringField('Minimum Duration', default='00:05:00')
    max_duration = StringField('Maximum Duration', default='00:15:00')
    published_since = StringField('Published Since', default=published_since)
    submit_field = SubmitField('Filter')

class SyncForm(FlaskForm):
    action = '/sync'
    submit_field = SubmitField('Sync')

async def fetch_videos(user_id, min_duration = None, max_duration = None, published_since = None):
    filters = {}

    if min_duration:
        filters['duration[>=]'] = f"'{min_duration}'"

    if max_duration:
        filters['duration[<=]'] = f"'{max_duration}'"

    if published_since:
        filters['published_at[>=]'] = f"'{published_since}'"

    videos_resource = f'{app.config["API_BASE"]}/videos/{user_id}'

    async with aiohttp.ClientSession() as session:
        async with session.get(videos_resource, params=filters) as response:
            videos = await response.json()
            return videos or []
    

async def sync_videos(user_id, token):
    videos_resource = f'{app.config["API_BASE"]}/videos/sync/{user_id}'
    headers = {'Authorization': token}

    async with aiohttp.ClientSession() as session:
        await session.post(videos_resource, headers=headers)

@app.route('/', methods=['GET', 'POST'])
async def root():
    if not is_logged_in():
        return redirect('/login')

    user_id = flask.session['user_info']['id']

    form = FilterForm()
    videos = await fetch_videos(
        user_id, form.min_duration.data, form.max_duration.data, form.published_since.data)
    return render_template('index.html', filter_form=form, sync_form=SyncForm(action='/sync'), videos=videos)


def redirect_uri():
    return f'{app.config["GOOGLE_REDIRECT_URI"]}/auth-code'

def is_logged_in():
    try:
        session_expiration = datetime.datetime.strptime(session.get('expires_at', ''), '%Y-%m-%dT%H:%M:%S')
        return datetime.datetime.utcnow() < session_expiration
    except ValueError:
        return False

@app.route('/login')
async def login():

    if is_logged_in():
        return redirect('/')

    config = app.config['GOOGLE_CONFIG']
    auth_uri = config['AUTH_URI']
    client_id = config['CLIENT_ID']
    scopes = "%20".join([
        "https://www.googleapis.com/auth/userinfo.profile",
        "https://www.googleapis.com/auth/youtube.readonly",
    ])
    google_redirect = redirect_uri()

    uri = f'{auth_uri}?client_id={client_id}&redirect_uri={google_redirect}&scope={scopes}&response_type=code'

    return redirect(uri)

@app.route('/auth-code')
async def auth_code():
    timestamp = datetime.datetime.utcnow()
    config = app.config['GOOGLE_CONFIG']
    token_uri = config['TOKEN_URI']
    token_params = {
        'code': request.args.get('code'),
        'client_id': config['CLIENT_ID'],
        'client_secret': config['CLIENT_SECRET'],
        'redirect_uri': redirect_uri(),
        'grant_type': 'authorization_code',
    }

    data = aiohttp.FormData(token_params)

    async with aiohttp.ClientSession() as session:
        headers = {}
        async with session.post(token_uri, data=data) as response:
            token_data = await response.json()
            flask.session['token'] = token_data
            headers['Authorization'] = f'Bearer {token_data["access_token"]}'

        if token_data:
            flask.session['expires_at'] = (
                timestamp + datetime.timedelta(seconds=token_data['expires_in'])
            ).strftime('%Y-%m-%dT%H:%M:%S')
            async with session.get('https://www.googleapis.com/oauth2/v1/userinfo', headers=headers) as response:
                user_info = await response.json()
                flask.session['user_info'] = user_info
                return redirect('/')
        else:
            return redirect('/login')

@app.route('/sync', methods=['POST'])
async def sync():
    if not is_logged_in():
        return redirect('/login')

    user_id = flask.session['user_info']['id']
    token = json.dumps(flask.session['token'])
    await sync_videos(user_id, token)

    return redirect('/')
