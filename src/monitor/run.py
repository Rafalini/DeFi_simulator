#!/usr/bin/env python
from web_app import app
import os, threading, time
from flask import request


if __name__ == "__main__":
    app.run(host='0.0.0.0', port=os.getenv("FLASK_SERVER_PORT"), debug=True)