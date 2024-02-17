import json
import os
from flask import Response, request
from web_app import app
# from bs4 import BeautifulSoup
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'

from web_app import ora

@app.route("/get-values")
def getCurrencies():
    return Response(json.dumps(ora.pricesTable), status=200, mimetype='application/json')

@app.route("/get-prices")
def getSymbols():
  # return Response(json.dumps(ora.getCurrentPrices()), status=200, mimetype='application/json')
  return Response(json.dumps(ora.getScheduledPrices()), status=200, mimetype='application/json')

@app.route("/set-modifier")
def setModifier():
    result = ora.setModifier(request.args.get('token'),request.args.get('value'))
    if ora.setModifier(request.args.get('token'),request.args.get('value')):
      return Response("New value: "+str(result), status=200)
    else:
      return Response("Error", status=400)
