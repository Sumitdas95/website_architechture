from django.contrib import admin
from .models import Topic, AccessRecord, Webpage, Company, Employee, Project

admin.site.register(Topic)
admin.site.register(AccessRecord)
admin.site.register(Webpage)
admin.site.register(Company)
admin.site.register(Employee)
admin.site.register(Project)