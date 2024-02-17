import os, requests, threading, time, csv
import pandas as pd

class Oracle:
    def __init__(self):
        self.scenarioFile = './web_app/scenario_const.csv'
        # self.scenarioFile = './web_app/scenario1.csv'
        self.checkpoint_path = "./weights"
        self.logFile = "log/oracle_log.csv"
        addr = os.getenv("AMM_SERVER_ADDR")
        port = int(os.getenv("AMM_SERVER_PORT"))
        self.ammUrl = "http://"+addr+":"+str(port)

        self.start_time = time.time()
        self.currencyIds = self.getCurrencies()
        self.pricesTable = self.getPricesAndAmountsFromCoinMarket()

        self.fieldNames = ["time", self.pricesTable[0]["symbol"]+","+self.pricesTable[1]["symbol"]+","+self.pricesTable[2]["symbol"]]

        f = open(self.logFile, "w")
        f.write("time_ora"+','+ self.pricesTable[0]["symbol"]+","+self.pricesTable[1]["symbol"]+","+self.pricesTable[2]["symbol"])
        f.write('\n')
        f.close()

        self.scenario = pd.read_csv(self.scenarioFile)

    def getCurrencies(self):
        url1 = self.ammUrl+"/get-currencies"
        ids = ""
        try:
            response1 = requests.get(url1)
            for entry in response1["currencies"]:
                ids += entry["id"]+","
        except Exception:
            pass
        return ids

    def getPricesAndAmountsFromCoinMarket(self):
        API_KEY = os.getenv("APIKEY")
        url = 'https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest'

        parameters = {
            'id': '1027,1518,5176',
            # 'limit': 5000
        }

        headers = {
            'Accepts': 'application/json',
            'X-CMC_PRO_API_KEY': API_KEY,
        }

        response = requests.get(url, headers=headers, params=parameters)
        data = response.json()
        usdPrices = []

        usdPrices.append({"symbol" : data["data"]["1027"]["symbol"], "usdprice" :  data["data"]["1027"]["quote"]["USD"]["price"], "modificator":0 })
        usdPrices.append({"symbol" : data["data"]["5176"]["symbol"], "usdprice" :  data["data"]["5176"]["quote"]["USD"]["price"], "modificator":0 })
        usdPrices.append({"symbol" : data["data"]["1518"]["symbol"], "usdprice" :  data["data"]["1518"]["quote"]["USD"]["price"], "modificator":0 })

        return usdPrices

    def saveStep(self):
        pricesTable = self.pricesTable
        elapsed_time = time.time() - self.start_time
        elapsed_milliseconds = int(elapsed_time * 1000)

        fieldnames = ["time_ora", pricesTable[0]["symbol"], pricesTable[1]["symbol"], pricesTable[2]["symbol"]]
        if elapsed_time < 60:
            with open(self.logFile, 'a', newline='') as csvfile:
                writer = csv.DictWriter(csvfile, fieldnames=fieldnames)

                if csvfile.tell() == 0:
                    writer.writeheader()

                data = {"time_ora":elapsed_milliseconds,
                        pricesTable[0]["symbol"]:pricesTable[0]["usdprice"]+pricesTable[0]["modificator"],
                        pricesTable[1]["symbol"]:pricesTable[1]["usdprice"]+pricesTable[0]["modificator"],
                        pricesTable[2]["symbol"]:pricesTable[2]["usdprice"]+pricesTable[0]["modificator"]}
                writer.writerow(data)
                csvfile.close()

    def getCurrentPrices(self):
        modifiedValues = []
        for entry in self.pricesTable:
            modifiedValues.append({"symbol":entry["symbol"], "usdprice":float(entry["usdprice"] + entry["modificator"])})
        self.saveStep()
        return modifiedValues

    def getScheduledPrices(self):
        condition = self.scenario['time_ms'] > (time.time() - self.start_time) * 1000
        match = self.scenario.loc[condition]

        if match.empty:
            df_reversed = self.scenario[::-1]
            match = df_reversed.loc[~condition].iloc[0]
        else:
            match = match.iloc[0]

        for entry in self.pricesTable:
            entry['usdprice'] = match[entry['symbol']]

        return self.getCurrentPrices()

    def setModifier(self, token, value):
        for entry in self.pricesTable:
            if entry["symbol"] == token:
                entry["modificator"] = value
                return entry["usdprice"] + entry["modificator"]
        return 0
