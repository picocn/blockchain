from django.shortcuts import render
from django.http import HttpResponse
from django.http import HttpResponseRedirect
from af import models
import json
import requests
# Create your views here.
def index(request):
    #return HttpResponse(u"欢迎光临 自强学堂!")
    return render(request,"layout.html")

def bankidx(request):
    cars=[{"id":1,"dealer":"福州永达", "model":"audi A6L", "factory":"Volkswagen", "amount":1000},
          {"id": 2, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
          {"id": 3, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
          {"id": 4, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
          {"id": 5, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},]
    return render(request,'bank.html',{"cars":cars})

def dealeridx(request):
    print("aaaaa")
    form = models.Order()
    cars = [{"id": 1, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
            {"id": 2, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
            {"id": 3, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
            {"id": 4, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000},
            {"id": 5, "dealer": "福州永达", "model": "audi A6L", "factory": "Volkswagen", "amount": 1000}, ]
    return render(request,'dealer.html',{"cars":cars, "form":form})

def dealer_makeorder(request, GET=None, POST=None):
    print("make order")
    if request.method=='GET':
        return HttpResponseRedirect('/dealer')
    elif request.method=='POST':
        print("POST:----------")
        url = "http://192.168.73.129:5000/registrar"
        data = '''{ "enrollId": "bob", "enrollSecret": "NOE63pEQbL25"}'''
        r=requests.post(url,data=data)
        print(r.text)
        return HttpResponseRedirect('/dealer')