from django.shortcuts import get_object_or_404, render

from .models import Bug

def index(request):
    bugs = Bug.objects.order_by('-created_at')[:5]
    return render(request, 'bugs/index.html', {'bugs': bugs})

def detail(request, bug_id):
    bug = get_object_or_404(Bug, pk=bug_id)
    return render(request, 'bugs/detail.html', {'bug': bug})
