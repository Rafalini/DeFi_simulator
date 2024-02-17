import json, math, hashlib, csv, binascii
from locale import currency
from datetime import datetime
import threading, time, os
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import serialization, hashes
from blockchainAdapter import BlockchainOrganizer

logFile = "log/amm_log.csv"
logTransFile = "log/amm_trans_log.csv"
fieldnames = ['time_amm','ETH_amount', 'XAUt_amount', 'MKR_amount']

class AmmClass:

    def __init__(self, filePath):
        self.transactionLog = []
        self.startTime = time.time()
        self.private_key = rsa.generate_private_key(
            public_exponent=65537,  # Commonly used value for the public exponent
            key_size=2048  # Size of the key in bits
        )

        self.public_key = self.private_key.public_key()

        self.pem_private = self.private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.TraditionalOpenSSL,
            encryption_algorithm=serialization.NoEncryption()
        )

        # Serialize the public key to PEM format
        self.pem_public = self.public_key.public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo
        )

        config_file = open(filePath)
        self.config = json.load(config_file)
        self.currencies = {}
        self.transactions = []
        self.pendingTransactions = []
        self.transactionCacheLimit = 8

        f = open(logFile, "w")
        f.write(','.join(fieldnames))
        f.write('\n')
        f.close()

        for entry in self.config["currencies"]:
            self.currencies[entry["short"]] = {"amount":entry["amount"], "minimal_part": entry["minimal_part"], "volume": 0}

        self.blockchain = BlockchainOrganizer()
        download_thread = threading.Thread(target=self.periodicSave, name="updater")
        download_thread.start()

    def getCurrencies(self):
        return self.currencies
    
    # def getAmountsDeprecated(self):
    #     self.saveStep()    
    #     amounts = {}

    #     btcSum, ethSum = 0,0
    #     n = 30
    #     if len(self.pendingTransactions)>n:
    #         for transaction in self.pendingTransactions[-n:]:
    #             if transaction["sender"] == "0xAMM":
    #                 if transaction["token"] == "BTC":
    #                     btcSum -= transaction["amount"]
    #                 if transaction["token"] == "ETH":
    #                     ethSum -= transaction["amount"] 
    #             if transaction["reciever"] == "0xAMM":
    #                 if transaction["token"] == "BTC":
    #                     btcSum += transaction["amount"]
    #                 if transaction["token"] == "ETH":
    #                     ethSum += transaction["amount"] 

    #     for currency in self.currencies:
    #         if currency == "BTC":
    #             amounts[currency] = self.currencies[currency]["amount"] - btcSum
    #         if currency == "ETH":
    #             amounts[currency] = self.currencies[currency]["amount"] - ethSum
    #     return amounts
    
    def getAmounts(self):
        amounts = []

        for currency in self.currencies:
            amounts.append({"symbol":currency,"amount":self.currencies[currency]["amount"]})    
        return amounts
    
    def requestedByConstantProduct(self, amount, token, exchangeToken):
        # minimal_part = 0.001 -> log10(minimal_part) = -3.0 -> round & abs -> 3 decimal numbers
        constK = self.currencies[exchangeToken]["amount"] * self.currencies[token]["amount"]

        deimalRoundingDigits = abs(round(math.log10(self.currencies[token]["minimal_part"])))
        requested = self.currencies[exchangeToken]["amount"] - round(
            constK / (amount + self.currencies[token]["amount"]), deimalRoundingDigits)
        return requested + 0.15
    
    def requestedByConstantSum(self, amount, token, exchangeToken):
        # minimal_part = 0.001 -> log10(minimal_part) = -3.0 -> round & abs -> 3 decimal numbers
        constK = self.currencies[exchangeToken]["amount"] + self.currencies[token]["amount"]
        if constK - (amount + self.currencies[token]["amount"]) > 0:
            requested = amount
        else:
            requested = self.currencies[token]["amount"]
        return requested

    # def proportional(self, amount, token, exchangeToken):
    #     # minimal_part = 0.001 -> log10(minimal_part) = -3.0 -> round & abs -> 3 decimal numbers
    #     deimalRoundingDigits = abs(round(math.log10(self.currencies[exchangeToken]["minimal_part"])))
    #     factor = self.currencies[exchangeToken]["amount"] / self.currencies[token]["amount"]
    #     requested = round(amount * factor, deimalRoundingDigits)
    #     return requested
    

    def getRates(self):
        response = {}    
        response["time"] = datetime.now().strftime("%H:%M:%S")
        list = self.getCurrentAmounts()
        print(list)
        for currency in list:
            rates = {}
            for referenceCurrency in list:
                if referenceCurrency != currency:
                    rates[referenceCurrency] = list[referenceCurrency] / list[currency]
                    # rates[referenceCurrency] = 1
            response[currency] = rates
        return response
    
    def parse_timestamp(self, timestamp_str):
        return datetime.strptime(timestamp_str, "%Y-%m-%dT%H:%M:%S.%fZ")

    def stop(self):
        # sorted_timestamps = sorted(self.transactionLog, key=self.parse_timestamp)
        # startTime = datetime.strptime(sorted_timestamps[-1], "%Y-%m-%dT%H:%M:%S.%fZ")
        with open(logTransFile, 'w', newline='') as csvfile:
            writer = csv.writer(csvfile)
            writer.writerow(['ms_elapsed'])
            
            for timestamp in self.transactionLog:
                # timestamp_datetime = datetime.strptime(timestamp, "%Y-%m-%dT%H:%M:%S.%fZ")
                # milliseconds_elapsed = int((startTime - timestamp_datetime).total_seconds() * 1000)
                writer.writerow([timestamp])


    def performTransaction(self, request):
        # self.transactionLog.append(request["TimeStamp"])

        time_difference = (time.time() - self.startTime) * 1000
        self.transactionLog.append(time_difference)

        # Convert the time difference to milliseconds
        amount = float(request["Amount"])
        token = request["Token"]
        print("exchange: " + request["Token"]+" for "+ request["Metadata"]["ExchangeToken"]+"|")
        exchangeToken = request["Metadata"]["ExchangeToken"]
        if token == "" or exchangeToken == "":
            return None
        if "Metadata" in request:
            exchangeRate = float(request["Metadata"]["ExchangeRate"])
            maxSlippage = float(request["Metadata"]["MaxSlippage"])
        requested = self.requestedByConstantProduct(amount, token, exchangeToken)
        # requested = self.requestedByConstantSum(amount, token, exchangeToken)

        # print("Exchange: "+token+" for "+exchangeToken+" ::: "+str(amount)+" for "+str(requested))

        returnTransaction = {}
        returnTransaction["TimeStamp"] = datetime.now().isoformat('T')
        returnTransaction["Sender"] = request["Reciever"]
        returnTransaction["Reciever"] = request["Sender"]
        if amount != 0:
            newRate = requested / amount
            if "Metadata" in request and abs( newRate / exchangeRate) <= maxSlippage and requested <= self.currencies[exchangeToken]["amount"]:
                #perform
                # print("slippage ok: "+str(abs(newRate - exchangeRate)))
                self.currencies[token]["amount"] += amount
                self.currencies[token]["volume"] += amount
                self.currencies[exchangeToken]["amount"] -= requested
                self.currencies[exchangeToken]["volume"] += requested

                self.transactions.insert(0, {"peer": request["Sender"], "from": token, "to": exchangeToken, "amountFrom": amount, "amountTo": requested})

                returnTransaction["Amount"] = str(requested)
                returnTransaction["Token"] = exchangeToken
            else:
                # print("slippage to low, rejected: "+str(abs(newRate / exchangeRate)))
                #reject
                returnTransaction["Amount"] = str(amount)
                returnTransaction["Token"] = token
        else:
            returnTransaction["Amount"] = str(amount)
            returnTransaction["Token"] = token
            
        transactionHash = hashlib.sha256(str(returnTransaction).encode('UTF-8')).digest()
        returnTransaction["TransactionHash"] = binascii.hexlify(transactionHash).decode('utf-8')
        signature = self.private_key.sign(
            transactionHash,
            padding.PKCS1v15(),
            hashes.SHA256())
        returnTransaction["SenderSignature"] = binascii.hexlify(signature).decode('utf-8')
        # print(self.currencies)
        return returnTransaction


    def lastTransactionChanges(self):
        changes = []
        if len(self.transactions) > 0:
            transaction = self.transactions[0]
            price = self.currencies[transaction["from"]]["amount"] / self.currencies[transaction["to"]]["amount"]
            lastPrice = (self.currencies[transaction["from"]]["amount"] - transaction["amountFrom"]) / (self.currencies[transaction["to"]]["amount"] + transaction["amountTo"])

            volumeChange = transaction["amountFrom"] / self.currencies[transaction["from"]]["volume"]
            priceChange = (price - lastPrice)/price

            changes.append({"currency":transaction["from"], "volume": round(self.currencies[transaction["from"]]["volume"],3), "volumeChange": round(volumeChange*100,5),
                            "price": round(price,3), "change": round(priceChange*100, 3)})

            price = self.currencies[transaction["to"]]["amount"] / self.currencies[transaction["from"]]["amount"]
            lastPrice = (self.currencies[transaction["to"]]["amount"] - transaction["amountTo"]) / (self.currencies[transaction["from"]]["amount"] + transaction["amountFrom"])

            volumeChange = transaction["amountTo"] / self.currencies[transaction["to"]]["volume"]
            priceChange = (lastPrice - price)/price

            changes.append({"currency":transaction["to"], "volume": round(self.currencies[transaction["to"]]["volume"],3), "volumeChange": round(volumeChange*100,5),
                            "price": round(price,3), "change": round(priceChange*100, 3)})
        return changes


    def getCurrentAmounts(self):
        amounts = {}
        for currency in self.currencies:
            amounts[currency] = self.currencies[currency]["amount"]
        return amounts
    
    def saveStep(self):
        f = open(logFile, "a")
        ls = self.getCurrentAmounts()
        data = [str(int((time.time() - self.startTime) * 1000)), str(ls["ETH"]), str(ls["XAUt"]), str(ls["MKR"])]
        f.write(','.join(data))
        f.write('\n')
        f.close()

    def periodicSave(self):
        while (time.time() - self.startTime) < int(os.getenv("DURATION")):
            self.saveStep()
            time.sleep(1)

    def addBlock(self, block):
        self.blockchain.addBlock(block)

    def getValidBlock(self):
        return self.blockchain.getValidBlock()

    def getAwaitingBlocks(self):
        return self.blockchain.getAwaitingBlocks()
    
    def getValidTransactions(self):
        return self.blockchain.getValidTransactions()