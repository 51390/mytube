import aiohttp
import datetime
import flask
import json
import os

from flask import Blueprint, Flask, render_template, redirect, request, current_app as app
from flask_wtf import FlaskForm
from wtforms import StringField, SubmitField

videos_blueprint = Blueprint('videos', __name__)


def is_logged_in():
    try:
        session_expiration = datetime.datetime.strptime(flask.session.get('expires_at', ''), '%Y-%m-%dT%H:%M:%S')
        return datetime.datetime.utcnow() < session_expiration
    except ValueError:
        return False


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


@videos_blueprint.route('/', methods=['GET', 'POST'])
async def root():
    if not is_logged_in():
        return redirect('/login')

    user_id = flask.session['user_info']['id']

    form = FilterForm()
    videos = await fetch_videos(
        user_id, form.min_duration.data, form.max_duration.data, form.published_since.data)
    return render_template('index.html', filter_form=form, sync_form=SyncForm(action='/sync'), videos=videos)


@videos_blueprint.route('/sync', methods=['POST'])
async def sync():
    if not is_logged_in():
        return redirect('/login')

    user_id = flask.session['user_info']['id']
    token = json.dumps(flask.session['token'])
    await sync_videos(user_id, token)

    return redirect('/')
