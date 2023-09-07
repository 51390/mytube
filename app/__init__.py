import aiohttp
import datetime
import flask
import json
import os

from flask import Flask, render_template, redirect, request, session
from flask_bootstrap import Bootstrap5
from flask_wtf import CSRFProtect
from flask_session import Session
from flask_sqlalchemy import SQLAlchemy
from wtforms import StringField, SubmitField

from app.login import login_blueprint
from app.videos import videos_blueprint

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
    db_password = os.environ.get('POSTGRES_PASSWORD', '')
    app = Flask(__name__, instance_relative_config=True)
    app.config.from_mapping(
        SECRET_KEY='dev',
        API_BASE='http://localhost:3000',
        PORT=port,
        GOOGLE_CONFIG_FILE="./client_secret.json",
        GOOGLE_REDIRECT_URI=f'{app_endpoint}:{port}',
        SQLALCHEMY_DATABASE_URI=f'postgresql://postgres:{db_password}@localhost:5432/yourtube_app',
        SESSION_TYPE='sqlalchemy',
    )

    db = SQLAlchemy(app)

    app.config.from_mapping(SESSION_SQLALCHEMY=db)
    app.config.from_prefixed_env(prefix='YOURTUBE')
    app.config.from_mapping(**load_credentials(app.config['GOOGLE_CONFIG_FILE']))

    bootstrap = Bootstrap5(app)
    csrf = CSRFProtect(app)
    Session(app)

    app.register_blueprint(login_blueprint)
    app.register_blueprint(videos_blueprint)

    return app, db, bootstrap, csrf

app, db, bootstrap, csrf = create_app()
