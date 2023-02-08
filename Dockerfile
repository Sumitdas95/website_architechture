FROM deliveroo/hopper-runner:1.9.9 as runner
FROM golang:1.19-buster

WORKDIR /app
ADD . .

RUN make install
RUN ln -s /app/bin/* /usr/bin

RUN echo 'export PS1="[$HOPPER_ECS_CLUSTER_NAME] $PS1"' >> /etc/profile.d/hopper_prompt.sh

WORKDIR /usr/bin/
COPY --from=runner /hopper-runner ./hopper-runner

RUN adduser app
USER app

ENTRYPOINT ["hopper-runner"]
