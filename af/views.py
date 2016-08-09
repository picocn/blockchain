from django.shortcuts import render
from django.http import HttpResponse


# Create your views here.
def index(request):
    #return HttpResponse(u"欢迎光临 自强学堂!")
    return render(request,"layout.html")

def bankidx(request):
    cars=[{"id":1,"dealer":"福州永达", "model":"audi A6L", "number":2, "amount":1000},
          {"id": 2, "dealer": "福州永达", "model": "audi A6L", "number": 2, "amount": 1000},
          {"id": 3, "dealer": "福州永达", "model": "audi A6L", "number": 2, "amount": 1000},
          {"id": 4, "dealer": "福州永达", "model": "audi A6L", "number": 2, "amount": 1000},
          {"id": 5, "dealer": "福州永达", "model": "audi A6L", "number": 2, "amount": 1000},]
    return render(request,'bank.html',{"cars":cars})