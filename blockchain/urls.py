"""blockchain URL Configuration

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/1.10/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  url(r'^$', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  url(r'^$', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.conf.urls import url, include
    2. Add a URL to urlpatterns:  url(r'^blog/', include('blog.urls'))
"""
from django.conf.urls import url
from django.contrib import admin
from af import views

urlpatterns = [
    url(r'^$', views.index),
    url(r'^bank/$', views.bankidx),
    url(r'^bank/grantloan/(?P<orderid>\w+\d+)/$', views.grantloan),
    url(r'^dealer/$', views.dealeridx),
    url(r'^dealer/makeorder/', views.dealer_makeorder),
    url(r'^dealer/orderlist/', views.dealer_orderlist),
    url(r'^admin/', admin.site.urls),
]
