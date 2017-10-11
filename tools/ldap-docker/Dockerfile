FROM debian:jessie

#############################################
# ApacheDS installation
#############################################

ENV APACHEDS_VERSION 2.0.0-M24
ENV APACHEDS_ARCH amd64

ENV APACHEDS_ARCHIVE apacheds-${APACHEDS_VERSION}-${APACHEDS_ARCH}.deb
ENV APACHEDS_DATA /var/lib/apacheds-${APACHEDS_VERSION}
ENV APACHEDS_USER apacheds
ENV APACHEDS_GROUP apacheds

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections \
    && apt-get update \
    && apt-get install -y ldap-utils procps openjdk-7-jre-headless curl \
    && curl http://www.eu.apache.org/dist//directory/apacheds/dist/${APACHEDS_VERSION}/${APACHEDS_ARCHIVE} > ${APACHEDS_ARCHIVE} \
    && dpkg -i ${APACHEDS_ARCHIVE} \
	&& rm ${APACHEDS_ARCHIVE}

#############################################
# ApacheDS bootstrap configuration
#############################################

ENV APACHEDS_INSTANCE default
ENV APACHEDS_BOOTSTRAP /bootstrap

ENV APACHEDS_INSTANCE_DIRECTORY ${APACHEDS_DATA}/${APACHEDS_INSTANCE}

WORKDIR /opt/apacheds-${APACHEDS_VERSION}/bin

ADD instance/* ${APACHEDS_BOOTSTRAP}/conf/
RUN mkdir ${APACHEDS_BOOTSTRAP}/cache \
    && mkdir ${APACHEDS_BOOTSTRAP}/run \
    && mkdir ${APACHEDS_BOOTSTRAP}/log \
    && mkdir ${APACHEDS_BOOTSTRAP}/partitions \
    && chown -R ${APACHEDS_USER}:${APACHEDS_GROUP} ${APACHEDS_BOOTSTRAP}

ADD data/ /
ADD scripts/bootstrap.sh /bootstrap.sh
RUN chown ${APACHEDS_USER}:${APACHEDS_GROUP} /bootstrap.sh \
    && chmod u+rx /bootstrap.sh && /bootstrap.sh

ADD scripts/run.sh /run.sh
RUN chown ${APACHEDS_USER}:${APACHEDS_GROUP} /run.sh \
    && chmod u+rx /run.sh

#############################################
# ApacheDS wrapper command
#############################################
CMD ["/run.sh"]
