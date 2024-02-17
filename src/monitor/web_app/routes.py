from web_app import amm, app, logger, oracleUrl
from flask import Response, request, render_template, stream_with_context

import json, time, requests, os

from datetime import datetime
timeInterval = 0.3

@app.route("/")
@app.route("/home")
def home():
    data = []
    for entry in amm.getCurrencies():
        data.append({"currency":entry})
    return render_template("index.html", changes=data)


@app.route("/get-currencies")
def getCurrencies():
    return Response(json.dumps(amm.getCurrencies()), status=200, mimetype='application/json')


def getChanges():
    client_ip = request.remote_addr
    try:
        while True:
            yield f"data:{json.dumps(amm.lastTransactionChanges())}\n\n"
            time.sleep(timeInterval)
    except GeneratorExit:
        logger.info("Client %s disconnected", client_ip)


def getCurrentAmounts():
    client_ip = request.remote_addr
    try:
        logger.info("Client %s connected", client_ip)
        while True:

            data = {}
            data["time"] = datetime.now().strftime("%H:%M:%S")

            try:
                data["amounts"] = amm.getAmounts()
                data["rates"] = amm.getRates()
                data["transactions"] = amm.getTransactions()
                data["prices"] = requests.get(oracleUrl+'/get-prices').json()
            except Exception as e:
                print(e)

            yield f"data:{json.dumps(data)}\n\n"
            time.sleep(timeInterval)
    except GeneratorExit:
        logger.info("Client %s disconnected", client_ip)


@app.route("/chart-data", methods=['GET'])
def chartData() -> Response:
    response = Response(stream_with_context(getCurrentAmounts()), mimetype="text/event-stream")
    response.headers["Cache-Control"] = "no-cache"
    response.headers["X-Accel-Buffering"] = "no"
    return response


@app.route("/changes-data", methods=['GET'])
def changesData() -> Response:
    response = Response(stream_with_context(getChanges()), mimetype="text/event-stream")
    response.headers["Cache-Control"] = "no-cache"
    response.headers["X-Accel-Buffering"] = "no"
    return response


@app.route("/get-rates", methods=['GET'])
def getRates() -> Response:
    """ {
            "BTC": {
                "BTC": 1.0,
                "ETH": 0.367309696969697
            },
            "ETH": {
                "BTC": 2.72249823037615,
                "ETH": 1.0
            }
        } 
    """
    return amm.getRates()