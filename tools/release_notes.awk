#!/usr/bin/env -S awk -f ${_} --

/^<!--/,/^-->/ { next }
/^## \[([0-9]+\.[0-9]+\.[0-9]+)\]\s*.*/ {
  if (!found) {
    found = 1
  } else {
    exit
  }
}
found { print }
