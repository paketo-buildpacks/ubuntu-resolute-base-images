# We want the minimal build image, as it will not be used for anything other than building the run image.
FROM ubuntu:resolute

ARG sources
ARG packages
ARG architecture
ARG package_args
