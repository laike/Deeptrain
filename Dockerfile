FROM python:3.9

ENV PYTHONUNBUFFERED 1
ENV DOCKER 1

MAINTAINER zmh-program

RUN mkdir /opt/Deeptrain
WORKDIR /opt/Deeptrain
ADD . /opt/Deeptrain

RUN apt-get update \
  && apt-get install ffmpeg libsm6 libxext6  -y

RUN  python3 -m pip install --upgrade pip \
  && pip3 install opencv-python-headless -i https://pypi.douban.com/simple/ \
  && pip3 install cryptography -i https://pypi.douban.com/simple/ \
  && pip3 install -r requirements.txt -i https://pypi.douban.com/simple/ \
  && python3 manage.py collectstatic --noinput
