FROM tasque/node:4.8-trusty

ONBUILD COPY package.json /app/
ONBUILD RUN npm install
ONBUILD COPY . /app

CMD [ "npm", "start" ]
