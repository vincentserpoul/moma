#!/bin/sh

# compile docker image for docker env
docker run --rm -v /media/vsm/Vertex3/workspace/go/src/github.com/vincentserpoul/moma:/src centurylink/golang-builder

# build docker image with the new compiled binary
docker build -t vincentserpoul/moma /media/vsm/Vertex3/workspace/go/src/github.com/vincentserpoul/moma

# save docker images to moma.tar
docker save vincentserpoul/moma > moma.tar

# rsync static files, templates and confdocker run -dit --restart=always -v /home/core/www/moma/templates:/templates --name moma -p 80:9000 --link redis:redisserver vincentserpoul/moma
rsync -arv /media/vsm/Vertex3/workspace/go/src/github.com/vincentserpoul/moma/config coreosmoma://home/core/www/moma/
rsync -arv /media/vsm/Vertex3/workspace/go/src/github.com/vincentserpoul/moma/templates coreosmoma://home/core/www/moma/
rsync /media/vsm/Vertex3/workspace/go/src/github.com/vincentserpoul/moma/moma.tar coreosmoma://home/core/

# import new docker image in coreosmoma
ssh coreosmoma "docker load < /home/core/moma.tar"
ssh coreosmoma "docker stop moma;docker rm moma;docker run -dit --restart=always -v /home/core/www/moma/templates:/templates -v /home/core/www/moma/config:/config --name moma -p 80 --link redis:redisserver vincentserpoul/moma"
