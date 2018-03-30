readTrusted 'Dockerfile'

pipeline {
  agent any

  stages {
    stage('Build and Test') {
      agent {
        docker 'golang:1.9.2'
      }

      steps {
        sh """
        go get github.com/golang/dep/cmd/dep

        mkdir -p /go/src/github.com/influxdata
        cp -a $WORKSPACE /go/src/github.com/influxdata/changelog

        cd /go/src/github.com/influxdata/changelog
        dep ensure -v -vendor-only
        go test -parallel=1 ./...
        """
      }
    }

    stage('Dockerfile') {
      steps {
        sh """
        docker build -t influxdata/changelog:build-${BUILD_NUMBER} .
        """
      }

      post {
        always {
          sh "docker rmi -f influxdata/changelog:build-${BUILD_NUMBER}"
        }
      }
    }
  }
}
