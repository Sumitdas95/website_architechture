from django.db import models
from datetime import date


class Topic(models.Model):
    top_name = models.CharField(max_length=265, unique=True)

    def __str__(self):
        return self.top_name


class Webpage(models.Model):
    topic = models.ForeignKey(Topic, on_delete= models.CASCADE)
    name = models.CharField(max_length=265, unique=True)
    url = models.URLField(unique=True)

    def __str__(self):
        return self.name


class AccessRecord(models.Model):
    name = models.ForeignKey(Webpage, on_delete=models.CASCADE)
    date = models.DateField

    def __str__(self):
        return self.date


class Company(models.Model):
    company_name = models.CharField(max_length=265, unique=True)
    number_of_employee = models.IntegerField(default=0)

    def __str__(self):
        return self.company_name


class Employee(models.Model):
    employee_name = models.CharField(max_length=265)
    company_name = models.ForeignKey(Company, on_delete=models.CASCADE)
    date_of_join = models.DateField(default=date.today)

    def __str__(self):
        return self.employee_name


class Project(models.Model):
    project_name = models.CharField(max_length=265, unique=True)
    # Many to Many relationships...
    employee_name = models.ManyToManyField('Employee')
    # One to One relationships...
    team_lead = models.OneToOneField('Employee', on_delete=models.CASCADE, related_name='team_lead', null=True)

    def __str__(self):
        return self.project_name
