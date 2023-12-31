import aiohttp
import datetime
import dateutil.parser
import flask
import humanize
import json
import jwt
import os

from flask import Blueprint, render_template, redirect, request, current_app as app
from flask_login import login_required
from flask_wtf import FlaskForm
from wtforms import StringField, SubmitField


videos_blueprint = Blueprint('videos', __name__)


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


@videos_blueprint.app_template_filter('humanize_ellapsed_time')
def humanize_ellapsed_time(time_string):
    delta = datetime.datetime.now(datetime.timezone.utc) - dateutil.parser.isoparse(time_string)
    return humanize.naturaldelta(delta)


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
    headers = {'user-token': token}

    async with aiohttp.ClientSession() as session:
        await session.post(videos_resource, headers=headers)


def jwt_access_token(token):
    key = app.config['JWT_TOKEN_SECRET']
    return jwt.encode({"token": json.dumps(token)}, key, algorithm='HS256')


@videos_blueprint.route('/', methods=['GET', 'POST'])
@login_required
async def root():
    user_id = flask.session['user_info']['id']

    form = FilterForm()
    videos = await fetch_videos(
        user_id, form.min_duration.data, form.max_duration.data, form.published_since.data)
    return render_template('index.html', filter_form=form, sync_form=SyncForm(action='/sync'), videos=videos)


@videos_blueprint.route('/sync', methods=['POST'])
@login_required
async def sync():
    user_id = flask.session['user_info']['id']
    token = jwt_access_token(flask.session['token'])
    await sync_videos(user_id, token)

    return redirect('/')
