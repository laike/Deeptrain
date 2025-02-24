"""
WSGI config for Deeptrain project.

It exposes the WSGI callable as a module-level variable named ``application``.

For more information on this file, see
https://docs.djangoproject.com/en/3.2/howto/deployment/wsgi/
"""

import os
import django
from django.core.handlers.wsgi import WSGIHandler
from Deeptrain import settings

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'Deeptrain.settings')


class WSGIApplication(WSGIHandler):
    pass


def get_wsgi_application():
    """
    The public interface to Django's WSGI support. Return a WSGI callable.

    Avoids making django.core.handlers.WSGIHandler a public API, in case the
    internal WSGI implementation changes or moves in the future.

    """
    django.setup(set_prefix=False)

    from utils.loop import loop
    from applications.application import appManager
    loop.start()

    if settings.IS_DEPLOYED:
        appManager.deploy_app()
    else:
        appManager.product_app()

    return WSGIApplication()


application = get_wsgi_application()
