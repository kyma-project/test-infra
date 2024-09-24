FROM python:3.12.6-slim

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app

COPY ./cmd/cloud-run/signifysecretrotator .
COPY ./cmd/cloud-run/signifysecretrotator/requirements.txt .

RUN pip3 install --upgrade -r requirements.txt && \
	apt-get install ca-certificates

CMD ["gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "--timeout", "0", "signifysecretrotator:app"]
