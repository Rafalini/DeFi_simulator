import requests

class AmmAdapter:
    def __init__(self, sourceUrl) -> None:
        self.sourceUrl = "http://"+sourceUrl

    def getCurrencies(self):
        try:
            return requests.get(self.sourceUrl+'/get-currencies', timeout=1).json()
        except Exception as e:
            # print(e)
            return []

    def lastTransactionChanges(self):
        try:
            return requests.get(self.sourceUrl+'/get-changes').json()
        except Exception as e:
            # print(e)
            return []

    def getAmounts(self):
        try:
            return requests.get(self.sourceUrl+'/get-amounts').json()
        except Exception as e:
            # print(e)
            return []
        
    def getRates(self):
        try:
            return requests.get(self.sourceUrl+'/get-rates').json()
        except Exception as e:
            # print(e)
            return []

    def getTransactions(self):
        try:    
            return requests.get(self.sourceUrl+'/get-transactions').json()
        except Exception as e:
            # print(e)
            return []