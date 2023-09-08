import  aiohttp
import datetime

from flask import Blueprint, redirect, request, session as flask_session, current_app as app
from flask_login import login_user

login_blueprint = Blueprint('login', __name__)


class User:
    def __init__(self, id):
        self._id = id

    def is_authenticated(self):
        try:
            print(flask_session.get('expires_at'))
            session_expiration = datetime.datetime.strptime(flask_session.get('expires_at', ''), '%Y-%m-%dT%H:%M:%S')
            return datetime.datetime.utcnow() < session_expiration
        except ValueError:
            return False

    def is_active(self):
        return self.is_authenticated()

    def is_anonymous(self):
        return False

    def get_id(self):
        return self._id


def redirect_uri():
    return f'{app.config["GOOGLE_REDIRECT_URI"]}/auth-code'


@login_blueprint.route('/login')
async def login():
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

@login_blueprint.route('/auth-code')
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
            flask_session['token'] = token_data
            headers['Authorization'] = f'Bearer {token_data["access_token"]}'

        if token_data:
            flask_session['expires_at'] = (
                timestamp + datetime.timedelta(seconds=token_data['expires_in'])
            ).strftime('%Y-%m-%dT%H:%M:%S')
            async with session.get('https://www.googleapis.com/oauth2/v1/userinfo', headers=headers) as response:
                user_info = await response.json()
                login_user(User(user_info['id']))
                flask_session['user_info'] = user_info
                return redirect('/')
        else:
            return redirect('/login')
