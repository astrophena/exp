from django.urls import path

from . import views

app_name = 'bugs'
urlpatterns = [
    path('', views.index, name='index'),
    path('<int:bug_id>/', views.detail, name='detail')
]
