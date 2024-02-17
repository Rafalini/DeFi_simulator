#!/usr/bin/env python
from web_app import app
import os

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=os.getenv("FLASK_ORACLE_PORT"), debug=True)