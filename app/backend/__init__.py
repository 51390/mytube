import aiohttp
import os

from flask import Flask, render_template
from flask_bootstrap import Bootstrap5
from flask_wtf import FlaskForm, CSRFProtect
from wtforms import StringField, SubmitField

def create_app():
    app = Flask(__name__, instance_relative_config=True)
    app.config.from_mapping(
        SECRET_KEY='dev',
        API_BASE='http://localhost:3000',
    )
    app.config.from_prefixed_env(prefix='YOURTUBE')

    try:
        os.makedirs(app.instance_path)
    except OSError:
        pass

    return app

app = create_app()
bootstrap = Bootstrap5(app)
csrf = CSRFProtect(app)

class FilterForm(FlaskForm):
    min_duration = StringField('Minimum Duration', default='05:00')
    max_duration = StringField('Maximum Duration', default='15:00')
    published_since = StringField('Published Since', default='2023-08-27')
    submit_field = SubmitField('Filter')

async def fetch_videos(min_duration = None, max_duration = None, published_since = None):
    filters = {}

    if min_duration:
        filters['duration[>=]'] = f"'{min_duration}'"

    if max_duration:
        filters['duration[<=]'] = f"'{max_duration}'"

    if published_since:
        filters['published_at[>=]'] = f"'{published_since}'"

    videos_resource = f'{app.config["API_BASE"]}/videos'

    async with aiohttp.ClientSession() as session:
        async with session.get(videos_resource, params=filters) as response:
            return await response.json()
    

@app.route('/', methods=['GET', 'POST'])
async def root():
    form = FilterForm()
    if form.validate_on_submit():
        print(f"{form.min_duration.data} / {form.max_duration.data} / {form.published_since.data}")
    videos = await fetch_videos(
        form.min_duration.data, form.max_duration.data, form.published_since.data)
    return render_template('index.html', form=form, videos=videos)
