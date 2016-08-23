# -----------------------------
# chaincode api
# ------------------------------
import requests
import json

def invoke(url, chaincodename, function, args, caller, id):
    url = url+"/chaincode"
    data = '''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                       "ctorMsg": {"function":"get_vehicles","args":["dealer"]},"secureContext":"dealer"},
                        "id": 30}'''
    chaincode = {"name":chaincodename}
    params = {"type": 1, "chaincodeID": chaincode}
    msg = {"function": function, "args": args}
    payload = {"jsonrpc": "2.0", "method": "invoke", "params":params, "ctorMsg": msg, "secureContext": caller, "id":id}
    payloadstr = json.dump(payload)
    rr = requests.post(url, data=payloadstr)


def restAPI(url, method, chaincodename, function, args, caller, id):
    url = url + "/chaincode"
    data = '''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                           "ctorMsg": {"function":"get_vehicles","args":["dealer"]},"secureContext":"dealer"},
                            "id": 30}'''
    chaincode = {"name": chaincodename}
    msg = {"function": function, "args": args}
    params = {"type": 1, "chaincodeID": chaincode, "ctorMsg": msg, "secureContext": caller}
    payload = {"jsonrpc": "2.0", "method": method, "params": params, "id": id}
    payloadstr = json.dump(payload)
    rr = requests.post(url, data=payloadstr)
