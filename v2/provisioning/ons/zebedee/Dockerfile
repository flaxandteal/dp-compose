FROM onsdigital/dp-concourse-tools-java

WORKDIR /zebedee

RUN mvn clean package dependency:copy-dependencies -Dmaven.test.skip=true -Dossindex.skip=true

CMD tail -f /dev/null