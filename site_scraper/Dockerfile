FROM python:3.9-alpine as base
FROM base as builder


COPY requirements.txt /requirements.txt
RUN \
  #apk update && \
  #apk add build-base && \
  #python -m venv /opt/venv && \
  pip install --upgrade pip && \
  pip install --user -r requirements.txt

FROM python:3-alpine
RUN pip install --user -r requirements.txt

FROM base
# copy only the dependencies installation from the 1st stage image
COPY --from=builder /root/.local /root/.local
WORKDIR /app
COPY . .

# update PATH environment variable
ENV PATH=/home/app/.local/bin:$PATH

ENTRYPOINT ["python", "main.py"]
#CMD ["/bin/sh"]

