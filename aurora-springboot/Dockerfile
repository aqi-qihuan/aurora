# Aurora 博客系统 - JDK 25 + Spring Boot 4.1.0-M4
FROM eclipse-temurin:25-jre-alpine

LABEL maintainer="Aurora Team"
LABEL description="Aurora Blog System - JDK 25 + Spring Boot 4.1.0-M4"
LABEL version="4.1.0-M4"

WORKDIR /app
VOLUME /tmp
ADD aurora-springboot-0.0.1.jar blog.jar

ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone

ENTRYPOINT ["java", \
  "-Xms96m", \
  "-Xmx192m", \
  "-XX:MetaspaceSize=128m", \
  "-XX:MaxMetaspaceSize=256m", \
  "-XX:ReservedCodeCacheSize=64m", \
  "-XX:+UseSerialGC", \
  "-Xss256k", \
  "-XX:+UseCompressedOops", \
  "-XX:+UseCompressedClassPointers", \
  "-XX:+UseStringDeduplication", \
  "-XX:+OptimizeStringConcat", \
  "-XX:+AlwaysPreTouch", \
  "-Djava.security.egd=file:/dev/./urandom", \
  "-Dspring.jmx.enabled=false", \
  "-Dspring.backgroundpreinitializer.ignore=true", \
  "-Dspring.profiles.active=prod", \
  "-jar", "blog.jar"]
