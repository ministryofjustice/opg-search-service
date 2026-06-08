# pact provider state endpoint
FROM python:3-alpine3.17@sha256:fc34b07ec97a4f288bc17083d288374a803dd59800399c76b977016c9fe5b8f2

WORKDIR /usr/src/app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ./api/* ./api/

CMD ["python", "./api/main.py"]
