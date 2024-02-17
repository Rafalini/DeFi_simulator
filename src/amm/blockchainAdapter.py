class BlockchainOrganizer:
        def __init__(self):
            self.blockchain = []
            self.awaitingBlocks = {}
            self.ageLimit = 8
            self.trustThreshold = self.ageLimit - 3

        def addBlock(self, block):            
            self.awaitingBlocks[block['Hash']] = {"block" : block, "prevoius" : block["PreviousHash"], "refNum": 0, "age": 0}

            for k in list(self.awaitingBlocks.keys()):
                if self.awaitingBlocks[k]["age"] > self.ageLimit:
                    del self.awaitingBlocks[k]

            for k in self.awaitingBlocks.keys():
                self.awaitingBlocks[k]["age"] += 1

            self.updateRefNumber(block["PreviousHash"])

            maxBlockKey = self.maxBlock()

            if self.awaitingBlocks[maxBlockKey]["refNum"] > self.trustThreshold:
                self.blockchain.append(self.awaitingBlocks[maxBlockKey]["block"])
                del self.awaitingBlocks[maxBlockKey]

            # print("adding: "+block["Hash"]+"  blockchain length: "+str(len(self.blockchain)))
            # print(self.awaitingBlocks.keys())


        def maxBlock(self):
            key = next(iter(self.awaitingBlocks.keys()), None)
            maxVal = self.awaitingBlocks[key]["refNum"]

            for k in self.awaitingBlocks.keys():
                if self.awaitingBlocks[k]["refNum"] > maxVal:
                    maxVal = self.awaitingBlocks[k]["refNum"]
                    key = k
            return key
            

        def updateRefNumber(self, hash):
            if hash not in self.awaitingBlocks:
                return
            
            self.awaitingBlocks[hash]["refNum"] += 1
            self.updateRefNumber(self.awaitingBlocks[hash]["prevoius"])


        def getValidTransactions(self):
            if len(self.blockchain)<1:
                return []
            return self.blockchain[-1]["Transactions"]
        
        def getValidBlock(self):
            if len(self.blockchain)<1:
                return {}
            return self.blockchain[-1]
        
        def getAwaitingBlocks(self):
            blocks = []
            for k in self.awaitingBlocks.keys():
                blocks.append({"block": self.awaitingBlocks[k]["block"], "age":self.awaitingBlocks[k]["age"], "refNum":self.awaitingBlocks[k]["refNum"]})
            return blocks
 