FROM local/unified-agent:latest

### Install python3.x
RUN apk add --no-cache python3 py3-pip
RUN pip3 install virtualenv

### Install Poetry (python)
### requires python3.X version matching the projects (defaults to python3.6)
### sed command sets the default selected python-executable used by poetry to be 'python3'
ENV POETRY_HOME ${WSS_USER_HOME}/.poetry
RUN curl -sSLO https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py && \
	sed -i 's/allowed_executa11bles = \["python", "python3"\]/allowed_executables = \["python3", "python"\]/g' get-poetry.py && \
	python3 get-poetry.py --yes --version 1.0.5 && \
	chown -R ${WSS_USER}:${WSS_GROUP} ${WSS_USER_HOME}/.poetry && \
	rm -rf get-poetry.py
ENV PATH ${WSS_USER_HOME}/.poetry/bin:${PATH}

COPY python-wss-unified-agent.config /wss
