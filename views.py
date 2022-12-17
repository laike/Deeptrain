from django.contrib import auth
from django.core.handlers.wsgi import WSGIRequest
from django.http import HttpResponse, Http404, JsonResponse
from django.shortcuts import render, redirect
from django.contrib.auth.decorators import login_required as _auth_login_required
from user.models import User
from user.forms import UserRegisterForm, UserLoginForm
from DjangoWebsite.settings import LOGIN_URL


def ajax_required(_decorate_exec: callable) -> callable:
    def _exec_function(request: WSGIRequest, *args, **kwargs) -> HttpResponse:
        if request.is_ajax():
            return _decorate_exec(
                request,
                *args, **kwargs,
            )
        else:
            raise Http404("404")

    return _exec_function


def login_required(_decorate_exec: callable) -> callable:
    @_auth_login_required
    def _exec_function(request: WSGIRequest, *args, **kwargs) -> HttpResponse:
        return _decorate_exec(
            request,
            request.user,
            *args, **kwargs
        )

    return _exec_function


def identity_required(level: int):
    def _decorate_(_decorate_exec: callable) -> callable:
        @login_required
        def _exec_(request: WSGIRequest, user, *args, **kwargs):
            if user.identity >= level:
                return _decorate_exec(
                    request, user,
                    *args, **kwargs
                )
            else:
                return render(request, 'permission.html')
        return _exec_
    return _decorate_


vip_required = identity_required(1)
admin_required = identity_required(2)
owner_required = identity_required(3)


def index(request: WSGIRequest) -> HttpResponse:
    return render(request, "index.html")


def login(request: WSGIRequest) -> HttpResponse:
    if request.POST:
        form = UserLoginForm(request.POST)
        if form.is_valid():
            auth.login(request, form.user)
            return JsonResponse({"success": True})
        else:
            return JsonResponse({"success": False, "reason": form.get_error()})
    else:
        if request.user.is_authenticated:
            return redirect("/home/")
        return render(request, 'login.html', {"form": UserLoginForm()})


def register(request: WSGIRequest) -> HttpResponse:
    if request.POST:
        form = UserRegisterForm(request.POST)
        form.country = getattr(request, "country")
        if form.is_valid():
            auth.login(request, form.user)
            return JsonResponse({"success": True})
        else:
            return JsonResponse({"success": False, "reason": form.get_error()})
    else:
        return render(request, 'register.html', {"form": UserRegisterForm()})


def logout(request: WSGIRequest) -> HttpResponse:
    auth.logout(request)
    return redirect(LOGIN_URL)


@login_required
def home(request: WSGIRequest, user) -> HttpResponse:
    if request.POST:
        _profile = request.POST.get("text")[:200].strip()
        user.profile.profile = _profile
        user.profile.save()
        return render(request, "home.html",
                      {"name": user.username, "profile": _profile, "id": user.real_identity})
    else:
        return render(request, "home.html",
                      {"name": user.username, "profile": user.profile.profile, "id": user.real_identity})


@login_required
def change(request: WSGIRequest, user):
    if request.POST:
        old, new = request.POST.get("old"), request.POST.get("new")
        user.check_password(old)
        user.set_password(new)
        user.save()
        logout(request)
    return redirect(LOGIN_URL)


def profile(request: WSGIRequest, uid) -> HttpResponse:
    query = User.objects.filter(id=uid)
    if query.exists():
        user = query.first()
        return render(request, "profile.html",
                      {"name": user.username, "profile": user.profile.profile, "id": user.real_identity})
    else:
        raise Http404("The user was not found on this server.")
