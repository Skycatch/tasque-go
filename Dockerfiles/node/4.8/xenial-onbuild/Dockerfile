FROM tasque/node:4.8-xenial

ONBUILD COPY package.json /app/
ONBUILD RUN npm install
ONBUILD COPY . /app

CMD [ "npm", "start" ]
