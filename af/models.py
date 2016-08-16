from django.db import models
from django import forms
import time
# Create your models here.

class Order(forms.Form):
    Factory = forms.ChoiceField(choices=(("faw","一汽轿车"),), required =True, label="汽车厂商")
    Model = forms.ChoiceField(choices=(("H7","红旗H7"),("L5","红旗L5"),("B90","奔腾B90"),("B70","奔腾B70"),), label="型号")
    Colour = forms.ChoiceField(choices=(('RED','红'),('WHITE','白'),('BLACK','黑')),label="颜色")
    Price = forms.DecimalField(label="贷款金额",max_digits=10, decimal_places=3)
    OrderID = forms.CharField(label="唯一订单号",initial="20160810")
    LoanBank = forms.ChoiceField(label="贷款银行",choices=(("cib","兴业银行"),("spdb","浦发银行"),("ccb","建设银行")))
    LoanContractID =forms.CharField(label='贷款合同编号')


class Transaction(models.Model):
    Participant = models.CharField(max_length=30)
    TransID = models.CharField(max_length=60)
    TransDate = models.DateField(auto_now_add=True)