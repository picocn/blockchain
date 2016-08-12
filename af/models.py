from django.db import models
from django import forms
# Create your models here.
"""
Factory := "\"Factory\":\"" + args[0] + "\", " // Variables to define the JSON
	Model := "\"Model\":\"" + args[1] + "\","
	Colour := "\"Colour\":\"" + args[2] + "\", "
	CarID := "\"CarID\":\"UNDEFINED\", "

	Price := "\"Price\":\"" + p + "\", "
	Dealer := "\"Dealer\":\"" + caller + "\", "
	Holder := "\"Holder\":\"" + argv[4] + "\", "
	Status := "\"Status\":\"INIT\", "
	OrderID := "\"OrderID\":\"" + caller + "_" + argv[5] + "\", "
	LoanContractID := "\"LoanContractID\":\"" + args[6] + "\""
"""
class Order(forms.Form):
    factory = forms.ChoiceField(choices=(("JAGULAR","捷豹路虎"),("法拉利","法拉利"),("保时捷","保时捷")),required =True, label="汽车厂商")
    Model = forms.CharField(label="型号")
    Colour = forms.ChoiceField(choices=(('RED','红'),('WHITE','白'),('黑','BLACK')),label="颜色")
    Price = forms.CharField(label="贷款金额")
    OrderID = forms.CharField(label="唯一订单号")
    LoanContractID =forms.CharField(label='贷款合同编号')
