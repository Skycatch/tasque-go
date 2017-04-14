FROM skycatch/tasque:6.10-alpine

ONBUILD COPY package.json /app/
ONBUILD RUN npm install
ONBUILD COPY . /app

ENV PKGNAME=graphicsmagick
ENV PKGVER=1.3.25
ENV PKGSOURCE=http://downloads.sourceforge.net/$PKGNAME/$PKGNAME/$PKGVER/GraphicsMagick-$PKGVER.tar.lz
    #wget name http://download.osgeo.org/libtiff/tiff-$TIFFVER.tar.gz && \
    #tar -xzvf tiff-$TIFFVER.tar.gz && \
    #cd tiff-$TIFFVER && \
    #./configure --prefix=/usr/local/libtiff-$TIFFVER && \
    #make && \
    #make install && \
    #ln -s libtiff-$TIFFVER /usr/local/libtiff && \
    # rm GraphicsMagick-$PKGVER.tar.lz && \

RUN apk add --update g++ \
    gcc \
    make \
    lzip \
    wget \
    libjpeg-turbo-dev \
    libpng-dev \
    libtool \
    tiff-dev \
    libgomp && \
    wget $PKGSOURCE && \
        lzip -d -c GraphicsMagick-$PKGVER.tar.lz | tar -xvf - && \
    cd GraphicsMagick-$PKGVER && \
        ./configure \
        --build=$CBUILD \
        --host=$CHOST \
        --prefix=/usr \
        --sysconfdir=/etc \
        --mandir=/usr/share/man \
        --infodir=/usr/share/info \
        --localstatedir=/var \
        --enable-shared \
        --disable-static \
        --with-modules \
        --with-threads \
        --with-gs-font-dir=/usr/share/fonts/Type1 \
        --with-quantum-depth=16 && \
        make && \
        make install && \
        cd ../ && \
        rm -rf GraphicsMagick-$PKGVER && \
        rm GraphicsMagick-$PKGVER.tar.lz && \
        apk del g++ \
                gcc \
                make \
                lzip \
                wget

CMD [ "npm", "start" ]
