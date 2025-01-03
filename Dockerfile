FROM alpine

COPY ./out/consumer /spectator/consumer

RUN wget "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz" -O ffmpeg.tar.xz
RUN tar -xf ffmpeg.tar.xz
ENV PATH=/root/.local/bin:$PATH
RUN echo 'export "PATH=$PATH:/ffmpeg/bin"' >> /etc/profile
RUN apk add gcompat

WORKDIR /spectator

CMD [ "./consumer" ]