# Artifact image: static files only. Served by the shared nginx in
# krapie/homeserver k8s/web, whose init container copies /site into it.
FROM busybox:1.38
COPY web/ /site/
