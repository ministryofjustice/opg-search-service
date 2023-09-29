# pact provider state endpoint
FROM python:3-alpine3.17

WORKDIR /usr/src/app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ./api/* ./api/

CMD ["python", "./api/main.py"]
