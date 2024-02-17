FROM python:3.11.6-bullseye

WORKDIR /usr/src/app

COPY . .

RUN pip install --no-cache-dir -r requirements.txt --break-system-packages

CMD [ "python", "-u", "run.py" ]