from django.contrib import admin

from .models import Bug


class BugAdmin(admin.ModelAdmin):
    list_display = ["title", "state", "owner", "created_at", "updated_at"]


admin.site.register(Bug, BugAdmin)
