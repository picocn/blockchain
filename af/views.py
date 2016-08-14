from django.shortcuts import render
from django.http import HttpResponse
from django.http import HttpResponseRedirect
from af import models
import json
import requests


# Create your views here.
def index(request):
    # return HttpResponse(u"欢迎光临 自强学堂!")
    return render(request, "layout.html")


def bankidx(request):
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "cib", "enrollSecret": "NOE63pEQbL25"}'''
    r = requests.post(url, data=data)
    print(r.text)
    url = "http://192.168.73.129:5000/chaincode"
    data='''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                   "ctorMsg": {"function":"get_vehicles","args":["cib"]},"secureContext":"cib"},
                    "id": 300}'''
    rr = requests.post(url, data=data)
    result= json.loads(rr.text)
    orders= json.loads(result["result"]["message"])
    print(orders)
    return render(request, 'bank.html', {"orders": orders})


def grantloan(request, orderid):
    print(orderid)
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "cib", "enrollSecret": "NOE63pEQbL25"}'''
    r = requests.post(url, data=data)
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "invoke","params": {"type": 1,"chaincodeID":{"name":"mycc"},
        "ctorMsg": {"function":"bank_confirm_order","args":["cib","'''
    data += orderid+ "\","+'''"faw"]},"secureContext":"cib"},
         "id": 300}'''
    rr = requests.post(url, data=data)
    print(data)
    return HttpResponseRedirect("/bank")

def dealeridx(request):
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "dealer", "enrollSecret": "jGlNl6ImkuDo"}'''
    r = requests.post(url, data=data)
    print(r.text)
    url = "http://192.168.73.129:5000/chaincode"
    data='''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                   "ctorMsg": {"function":"get_vehicles","args":["dealer"]},"secureContext":"dealer"},
                    "id": 30}'''
    rr = requests.post(url, data=data)
    result= json.loads(rr.text)
    orders= json.loads(result["result"]["message"])
    print(type(orders))
    form = models.Order()
    # cars = [{"id": 1, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 2, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 3, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 4, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 5, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000}, ]
    return render(request, 'dealer.html', {"orders": orders, "form": form})


def dealer_makeorder(request, GET=None, POST=None):
    print("make order")
    if request.method == 'GET':
        return HttpResponseRedirect('/dealer')
    elif request.method == 'POST':

        print("POST:----------")
        form = models.Order(request.POST)
        if not form.is_valid():
            print("form is not valid")
            return HttpResponseRedirect('/dealer')
        url = "http://192.168.73.129:5000/registrar"
        data = '''{ "enrollId": "dealer", "enrollSecret": "jGlNl6ImkuDo"}'''
        r = requests.post(url, data=data)
        print(r.text)
        url = "http://192.168.73.129:5000/chaincode"
        #if form.is_valid():
        data = '''{
               "jsonrpc": "2.0",
               "method": "invoke",
               "params": {
                   "type": 1,
                   "chaincodeID":{
                       "name":"mycc"
                   },
                   "ctorMsg": {
                       "function":"create_Order",
                       "args":['''
        args = "\"dealer\", \"" + form.cleaned_data["Factory"] + "\", \"" + form.cleaned_data["Model"]\
                   + "\", \"" + form.cleaned_data["Colour"] + "\", \"" \
                   + str(form.cleaned_data["Price"]) + "\", \"" + form.cleaned_data["OrderID"]\
                   + "\", \"" + form.cleaned_data["LoanBank"] + "\", \"" \
                   + form.cleaned_data["LoanContractID"] + "\" ]"
        data = data + args + '''},
                "secureContext":"dealer"
                     },
                    "id": 3
                }'''
        print(data)

        rr = requests.post(url, data=data)
        #x = json.loads(rr.text)
        print(rr.text)
        return HttpResponseRedirect('/dealer')


def dealer_orderlist(request):
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "dealer", "enrollSecret": "jGlNl6ImkuDo"}'''
    r = requests.post(url, data=data)
    print(r.text)
    url = "http://192.168.73.129:5000/chaincode"
    data='''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                   "ctorMsg": {"function":"get_vehicles","args":["dealer"]},"secureContext":"dealer"},
                    "id": 30}'''
    rr = requests.post(url, data=data)
    result= json.loads(rr.text)
    orders= json.loads(result["result"]["message"])
    print(type(orders))
    form = models.Order()
    # cars = [{"id": 1, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 2, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 3, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 4, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 5, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000}, ]
    return render(request, 'dealer_list.html', {"orders": orders, "form": form})