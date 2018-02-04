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
        mkdir -p /go/src/github.com/influxdata
        cp -a $WORKSPACE /go/src/github.com/influxdata/changelog

        cd /go/src/github.com/influxdata/changelog
        go get -d -t ./...
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
