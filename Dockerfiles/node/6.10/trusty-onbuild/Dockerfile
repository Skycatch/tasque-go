FROM tasque/node:6.10-trusty

ONBUILD COPY package.json /app/
ONBUILD RUN npm install
ONBUILD COPY . /app

CMD [ "npm", "start" ]
