#!/usr/bin/env python
import os
import sys
import logging, threading
from flask import Flask
from web_app.ammAdapter import AmmAdapter

logging.basicConfig(stream=sys.stdout, level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)
app = Flask(__name__)

addr = os.getenv("AMM_SERVER_ADDR")
port = int(os.getenv("AMM_SERVER_PORT"))

amm = AmmAdapter(addr+":"+str(port))

addr = os.getenv("ORACLE_SERVER_ADDR")
port = int(os.getenv("ORACLE_SERVER_PORT"))

oracleUrl =  "http://"+addr+":"+str(port)

from web_app import routes