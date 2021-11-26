from django.views import generic
from django.contrib.auth.decorators import login_required

from .models import Bug

class IndexView(generic.ListView):
    template_name = "bugs/index.html"
    context_object_name = "bugs"

    def get_queryset(self):
        """Return all bugs."""
        return Bug.objects.order_by("created_at")

class DetailView(generic.DetailView):
    model = Bug
    template_name = "bugs/detail.html"
