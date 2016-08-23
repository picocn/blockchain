from django.db import models
from django import forms
import time
import re
# Create your models here.

class Order(forms.Form):
    Factory = forms.ChoiceField(choices=(("faw","一汽轿车"),), required =True, label="汽车厂商")
    Model = forms.ChoiceField(choices=(("H7","红旗H7"),("L5","红旗L5"),("B90","奔腾B90"),("B70","奔腾B70"),), label="型号")
    Colour = forms.ChoiceField(choices=(('RED','红'),('WHITE','白'),('BLACK','黑')),label="颜色")
    Price = forms.DecimalField(label="贷款金额",max_digits=10, decimal_places=3)
    OrderID = forms.CharField(label="唯一订单号")
    LoanBank = forms.ChoiceField(label="贷款银行",choices=(("cib","兴业银行"),("spdb","浦发银行"),("ccb","建设银行")))
    LoanContractID =forms.CharField(label='贷款合同编号')

    def clean_OrderID(self):
        orderid = self.cleaned_data.get("OrderID", None)
        if orderid is None:
            return self.initial.get("OrderID", time.strftime("%Y%m%d%H%M%S"))
        pattern = re.compile(r"[a-zA-Z0-9_]+")
        if pattern.match(orderid):
            pass
        else:
            msg = "订单编号仅支持字母数字和_-的组合"
            self._errors["OrderID"] = self.error_class([msg])
        self.OrderID = orderid
        return orderid


class Transaction(models.Model):
    Participant = models.CharField(max_length=40)
    Action = models.CharField(max_length=40)
    TransID = models.CharField(max_length=60)
    TransDate = models.DateTimeField(auto_now_add=True)