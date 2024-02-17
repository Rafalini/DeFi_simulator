from re import A
from twisted.internet import reactor, endpoints
from twisted.web.server import Site
from twisted.web.resource import Resource
from twisted.internet.protocol import DatagramProtocol
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives import serialization
import socket, json, sys, threading
import logging, os, base64, binascii

logging.basicConfig(stream=sys.stdout, level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

from ammDefinition import AmmClass

amm = AmmClass('config.json')
# pastTransactions = []
counter = 0
TRANS_MCAST_ADRR = os.getenv("TRANSACTION_BROADCAST").split(":")[0]
TRANS_MCAST_PORT = int(os.getenv("TRANSACTION_BROADCAST").split(":")[1])

ammAdrr = os.getenv("AMM_SERVER_ADDR")
NODE_MCAST_ADRR = os.getenv("NODE_BROADCAST").split(":")[0]
NODE_MCAST_PORT = os.getenv("NODE_BROADCAST").split(":")[1]
MULTICAST_TTL = 1

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
sock.setsockopt(socket.IPPROTO_IP, socket.IP_MULTICAST_TTL, MULTICAST_TTL)

def shutdownAmm(amm):
    amm.stop()
    reactor.stop()

class BaseResource(Resource):
    isLeaf = True
    def setHeaders(self, request):
        request.setHeader(b'Content-Type', b'application/json')
        request.setResponseCode(200)
        global counter
        counter += 1
        # logger.info("req cont %s from %s", counter, request.path)
        # logger.info("Client %s connected to %s", request.getClientAddress(), request.path)
        

class GetCurrencies(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        return json.dumps(amm.getCurrencies()).encode('utf-8')
        # return json.dumps([]).encode('utf-8')

class GetChanges(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        return json.dumps(amm.lastTransactionChanges()).encode('utf-8')

class GetAmounts(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        return json.dumps(amm.getAmounts()).encode('utf-8')

class GetRates(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        return json.dumps(amm.getRates()).encode('utf-8')

class GetTransactions(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        return json.dumps(amm.getValidTransactions()).encode('utf-8')
    
class GetValidBlock(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        return json.dumps(amm.getValidBlock()).encode('utf-8')
    
class GetAwaitingBlocks(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        return json.dumps(amm.getAwaitingBlocks()).encode('utf-8')
    
class GetPublicKey(BaseResource):
    def render_GET(self, request):
        self.setHeaders(request)
        request.responseHeaders.addRawHeader(b"content-type", b"application/x-pem-file")

        return amm.public_key.public_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PublicFormat.PKCS1
            )


class MyRootResource(Resource):
    def __init__(self):
        Resource.__init__(self)
        self.putChild(b"get-currencies", GetCurrencies())
        self.putChild(b"get-changes", GetChanges()) #tofix
        self.putChild(b"get-amounts", GetAmounts())
        self.putChild(b"get-rates", GetRates())
        self.putChild(b"get-transactions", GetTransactions()) #softFix?
        self.putChild(b"get-public-key", GetPublicKey())
        self.putChild(b"get-valid-block", GetValidBlock())
        self.putChild(b"get-awaiting-blocks", GetAwaitingBlocks())

class NodeMulticastListener(DatagramProtocol):
    def startProtocol(self):
        # Specify the multicast address to listen to
        self.transport.joinGroup(NODE_MCAST_ADRR)

    def datagramReceived(self, data, addr):
        # Handle the incoming multicast UDP datagram here
        logger.info("Received Node UDP data: %d [b] from %s", len(data), addr)
        # print(data.decode('utf-8'))
        block = json.loads(data.decode('utf-8'))
        amm.addBlock(block)
        justValidatedTransactions = [x for x in amm.getValidTransactions() if x["Sender"] != ammAdrr]

        for trans in justValidatedTransactions:
            resulttransaction = amm.performTransaction(trans)
            if resulttransaction is not None:
                sock.sendto(bytes(json.dumps(resulttransaction), 'utf-8'), (TRANS_MCAST_ADRR, TRANS_MCAST_PORT))


class TransactionMulticastListener(DatagramProtocol):
    def startProtocol(self):
        self.transport.joinGroup(NODE_MCAST_ADRR)

    def datagramReceived(self, data, addr):
        pass

reactor.listenMulticast(int(NODE_MCAST_PORT), NodeMulticastListener())
reactor.listenMulticast(int(TRANS_MCAST_PORT), TransactionMulticastListener())

root = MyRootResource()
site = Site(root)

tcp_endpoint = endpoints.TCP4ServerEndpoint(reactor, 5000)
tcp_endpoint.listen(site)
tcp_endpoint2 = endpoints.TCP4ServerEndpoint(reactor, 5005)
tcp_endpoint2.listen(site)
reactor.callLater(int(os.getenv("DURATION"))+5, shutdownAmm, amm)
reactor.run()