FROM local/unified-agent:latest

### Install python3.x
RUN apk add --no-cache python3 py3-pip

## Install Conda (python)
USER ${WSS_USER}
RUN cd ${WSS_USER_HOME} && \
    wget https://repo.anaconda.com/archive/Anaconda3-2021.05-Linux-x86_64.sh && \
    bash Anaconda3-2021.05-Linux-x86_64.sh -b && \
    rm Anaconda3-2021.05-Linux-x86_64.sh

USER root
RUN echo '#!/usr/bin/env bash' >> /usr/bin/conda && \
    echo 'source ${WSS_USER_HOME}/anaconda3/etc/profile.d/conda.sh' >> /usr/bin/conda && \
    echo '${WSS_USER_HOME}/anaconda3/bin/conda "$@"' >> /usr/bin/conda && \
    chmod +x /usr/bin/conda

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
