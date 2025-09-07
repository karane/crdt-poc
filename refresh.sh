#!/bin/bash

SLIDES=$1

pandoc $SLIDES/slides.md -t beamer -o output/$SLIDES.pdf -V theme=Berlin
open output/$SLIDES.pdf
