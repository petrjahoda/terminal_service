FROM alpine:latest as build
RUN apk add tzdata
FROM scratch as final
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
ADD /linux /
CMD ["/terminal_service"]