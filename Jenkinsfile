pipeline {
    def app
    agent { dockerfile true }
    stages {

    stage('Clone repository') {
        /* Let's make sure we have the repository cloned to our workspace */

        git clone 'https://github.com/ValeryPiashchynski/TaskManager.git'
    }

    stage('Build image') {
        /* This builds the actual image; synonymous to
         * docker build on the command line */

        app = docker.build("TaskManager/gateway/main")
    }


        }
    }
}