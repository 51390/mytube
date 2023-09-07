import aiohttp
import datetime
import flask
import json
import os

from flask_bootstrap import Bootstrap5
from flask_wtf import CSRFProtect
from flask_session import Session
from flask_sqlalchemy import SQLAlchemy

from app.login import login_blueprint
from app.videos import videos_blueprint

def load_credentials():
    return {
        "GOOGLE_CONFIG": {
            "CLIENT_ID": os.environ.get('CLIENT_ID'),
            "CLIENT_SECRET": os.environ.get('CLIENT_SECRET'),
            "PROJECT_ID": os.environ.get('PROJECT_ID'),
            "AUTH_URI": os.environ.get('AUTH_URI'),
            "TOKEN_URI": os.environ.get('TOKEN_URI'),
            "AUTH_PROVIDER_X509_CERT_URL": os.environ.get('AUTH_PROVIDER_X509_CERT_URL'),
            "REDIREC_URIS": os.environ.get('REDIREC_URIS'),
        }
    }


def create_app(app_endpoint='http://localhost', port=5000):
    db_password = os.environ.get('POSTGRES_PASSWORD', '')
    app = flask.Flask(__name__, instance_relative_config=True)
    app.config.from_mapping(
        SECRET_KEY='dev',
        API_BASE='http://localhost:3000',
        PORT=port,
        GOOGLE_REDIRECT_URI=f'{app_endpoint}:{port}',
        SQLALCHEMY_DATABASE_URI=f'postgresql://postgres:{db_password}@localhost:5432/yourtube_app',
        SESSION_TYPE='sqlalchemy',
    )

    db = SQLAlchemy(app)

    app.config.from_mapping(SESSION_SQLALCHEMY=db)
    app.config.from_prefixed_env(prefix='MYTUBE')
    app.config.from_mapping(**load_credentials())

    bootstrap = Bootstrap5(app)
    csrf = CSRFProtect(app)
    Session(app)

    app.register_blueprint(login_blueprint)
    app.register_blueprint(videos_blueprint)

    return app, db, bootstrap, csrf

app, db, bootstrap, csrf = create_app()
