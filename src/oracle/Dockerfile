FROM tensorflow/tensorflow:2.13.0

WORKDIR /usr/src/app

COPY . .

RUN pip install --no-cache-dir -r requirements.txt

CMD [ "python", "-u", "./run.py" ]