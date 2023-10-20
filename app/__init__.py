import aiohttp
import datetime
import flask
import json
import os

from flask_bootstrap import Bootstrap5
from flask_login import LoginManager
from flask_wtf import CSRFProtect
from flask_session import Session
from flask_sqlalchemy import SQLAlchemy

from app.login import login_blueprint, login, User
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


def create_app():
    app_endpoint = os.environ.get('APP_ENDPOINT', 'http://localhost')
    app_port = int(os.environ.get('APP_PORT', 5000))
    db_password = os.environ.get('POSTGRES_PASSWORD', '')
    db_user = os.environ.get('POSTGRES_USER', 'postgres')
    db_host = os.environ.get('POSTGRES_HOST', 'localhost')
    api_base = os.environ.get('API_BASE', 'http://localhost:3000')

    app = flask.Flask(__name__, instance_relative_config=True)
    app.config.from_mapping(
        SECRET_KEY='dev',
        JWT_TOKEN_SECRET=os.environ.get('JWT_TOKEN_SECRET'),
        API_BASE=api_base,
        PORT=app_port,
        GOOGLE_REDIRECT_URI=f'{app_endpoint}:{app_port}',
        SQLALCHEMY_DATABASE_URI=f'postgresql://{db_user}:{db_password}@{db_host}:5432/mytube_app',
        SESSION_TYPE='sqlalchemy',
    )

    db = SQLAlchemy(app)

    app.config.from_mapping(SESSION_SQLALCHEMY=db)
    app.config.from_prefixed_env(prefix='MYTUBE')
    app.config.from_mapping(**load_credentials())

    bootstrap = Bootstrap5(app)
    login_manager = LoginManager()
    login_manager.init_app(app)
    csrf = CSRFProtect(app)
    Session(app)

    app.register_blueprint(login_blueprint)
    app.register_blueprint(videos_blueprint)

    login_manager.login_view = 'login.login'

    return app, db, bootstrap, csrf, login_manager

app, db, bootstrap, csrf, login_manager = create_app()


@login_manager.user_loader
def load_user(user_id):
    user = User(user_id)
    if user.is_authenticated():
        return user
    else:
        return None
