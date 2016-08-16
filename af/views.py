from django.shortcuts import render
from django.http import HttpResponse
from django.http import HttpResponseRedirect
from af import models
import json
import requests


# Create your views here.
def index(request):
    # return HttpResponse(u"欢迎光临 自强学堂!")
    return render(request, "layout.html",{"participant":"Participant"})


def status(request):
    url="http://192.168.73.129:5000/chain"
    r = requests.get(url)
    chain = json.loads(r.text)
    return render(request,"status.html",{"chain":chain, "participant":"Participant"})


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
    return render(request, 'bank.html', {"orders": orders, "participant":"兴业银行"})

def returnloan(request,orderid):
    print(orderid)
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "cib", "enrollSecret": "NOE63pEQbL25"}'''
    r = requests.post(url, data=data)
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "invoke","params": {"type": 1,"chaincodeID":{"name":"mycc"},
            "ctorMsg": {"function":"update_state_repayment","args":["cib","'''
    data += orderid + "\"" + ''']},"secureContext":"cib"},
             "id": 302}'''
    rr = requests.post(url, data=data)
    print(data)
    return HttpResponseRedirect("/bank/loanlist")

def bank_orderlist(request):
    url = "http://192.168.73.129:5000/chaincode"
    data='''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                   "ctorMsg": {"function":"get_all_vehicles","args":["cib"]},"secureContext":"cib"},
                    "id": 300}'''
    rr = requests.post(url, data=data)
    result= json.loads(rr.text)
    orders= json.loads(result["result"]["message"])
    print(orders)
    return render(request, 'bank_orderlist.html', {"orders": orders, "participant":"兴业银行"})

def bank_loanlist(request):
    url = "http://192.168.73.129:5000/chaincode"
    data='''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                   "ctorMsg": {"function":"get_repay_vehicles","args":["cib"]},"secureContext":"cib"},
                    "id": 300}'''
    rr = requests.post(url, data=data)
    result= json.loads(rr.text)
    orders= json.loads(result["result"]["message"])
    print(orders)
    return render(request, 'bank_loanlist.html', {"orders": orders, "participant":"兴业银行"})


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
    return render(request, 'dealer.html', {"orders": orders, "form": form, "participant":"经销商"})


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
        #print(data)

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
    allorder = False
    # cars = [{"id": 1, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 2, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 3, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 4, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
    #         {"id": 5, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000}, ]
    return render(request, 'dealer_list.html', {"orders": orders, "form": form, "allorder":allorder, "participant":"经销商"})


def dealer_allorderlist(request):
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "dealer", "enrollSecret": "jGlNl6ImkuDo"}'''
    r = requests.post(url, data=data)
    print(r.text)
    url = "http://192.168.73.129:5000/chaincode"
    data='''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                   "ctorMsg": {"function":"get_all_vehicles","args":["dealer"]},"secureContext":"dealer"},
                    "id": 30}'''
    rr = requests.post(url, data=data)
    result= json.loads(rr.text)
    orders= json.loads(result["result"]["message"])
    print(type(orders))
    form = models.Order()
    allorder=True
    #menu = '''<li ><a href="/dealer/orderlist">查看已提交订单</a></li>
    #        <li class="active"><a href="/dealer/allorderlist">全部订单<span class="sr-only">(current)</span></a></li>'''
    return render(request, 'dealer_list.html', {"orders": orders, "form": form, "allorder":allorder, "participant":"经销商"})

def logistics(request):
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "cosco", "enrollSecret": "NOE63pEQbL25"}'''
    r = requests.post(url, data=data)
    print(r.text)
    url = "http://192.168.73.129:5000/chaincode"
    data='''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                   "ctorMsg": {"function":"get_vehicles","args":["cosco"]},"secureContext":"cosco"},
                    "id": 222}'''
    rr = requests.post(url, data=data)
    result= json.loads(rr.text)
    orders= json.loads(result["result"]["message"])
    print(orders)
    return render(request, 'logistics.html', {"orders": orders, "participant":"中远物流"})


def logistics_updategeo(request,orderid):
    newgeo = request.GET.get("newgo")
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "invoke","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                               "ctorMsg": {"function":"update_loc","args":["cosco","'''
    data += orderid
    data += "\",\""
    data += newgeo
    data += '''"]},"secureContext":"cosco"},
                                "id": 123}'''
    print(data)
    rr = requests.post(url, data=data)
    #result = json.loads(rr.text)
    #orders = json.loads(result["result"]["message"])
    # print(orders)
    return HttpResponseRedirect("/logistics")


def logistics_deliver(request, orderid):
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "invoke","params": {"type": 1,"chaincodeID":{"name":"mycc"},
        "ctorMsg": {"function":"logistics_deliver","args":["cosco","'''
    data += orderid+ "\","+'''"dealer"]},"secureContext":"cosco"},
         "id": 300}'''
    rr = requests.post(url, data=data)
    #print(data)
    return HttpResponseRedirect("/logistics")


def logistics_geo(request):
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                           "ctorMsg": {"function":"get_vehicles","args":["cosco"]},"secureContext":"cosco"},
                            "id": 222}'''
    rr = requests.post(url, data=data)
    result = json.loads(rr.text)
    orders = json.loads(result["result"]["message"])
    # print(orders)
    return render(request, 'logisticsgeo.html', {"orders": orders, "participant":"中远物流"})


def manufacturer(request):
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "faw", "enrollSecret": "jGlNl6ImkuDo"}'''
    r = requests.post(url, data=data)
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                       "ctorMsg": {"function":"get_vehicles","args":["faw"]},"secureContext":"faw"},
                        "id": 222}'''
    rr = requests.post(url, data=data)
    result = json.loads(rr.text)
    orders = json.loads(result["result"]["message"])
    #print(orders)
    return render(request, 'manufacturer.html', {"orders": orders, "participant":"一汽轿车"})


def manufacturer_orderlist(request):
    url = "http://192.168.73.129:5000/registrar"
    data = '''{ "enrollId": "faw", "enrollSecret": "jGlNl6ImkuDo"}'''
    r = requests.post(url, data=data)
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "query","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                       "ctorMsg": {"function":"get_all_vehicles","args":["faw"]},"secureContext":"faw"},
                        "id": 222}'''
    rr = requests.post(url, data=data)
    result = json.loads(rr.text)
    orders = json.loads(result["result"]["message"])
    #print(orders)
    return render(request, 'manufacturer_orderlist.html', {"orders": orders, "participant":"一汽轿车"})

def manufacturer_deliver(request,orderid):
    carid = request.GET.get("carid")
    logistics_name = request.GET.get("logistics")
    url = "http://192.168.73.129:5000/chaincode"
    data = '''{"jsonrpc": "2.0","method": "invoke","params": {"type": 1,"chaincodeID":{"name":"mycc"},
                           "ctorMsg": {"function":"manufacturer_deliver","args":["faw","'''
    data += orderid
    data+= "\",\""
    data+= logistics_name
    data+= "\",\""
    data+=carid
    data+='''"]},"secureContext":"faw"},
                            "id": 222}'''
    rr = requests.post(url, data=data)
    #result = json.loads(rr.text)
    #orders = json.loads(result["result"]["message"])
    # print(orders)
    return HttpResponseRedirect("/manufacturer")