#!/usr/bin/env python
import os
import sys
import json
import logging
from flask import Flask
from web_app.oracle import Oracle

logging.basicConfig(stream=sys.stdout, level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)
app = Flask(__name__)
ora = Oracle()

from web_app import routes