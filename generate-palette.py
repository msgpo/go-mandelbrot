#!/usr/bin/env python3

import argparse
import numpy as np

from scipy.interpolate import PchipInterpolator

# https://stackoverflow.com/a/25816111
knots = [0.0, 0.16, 0.42, 0.6425, 0.8575, 1]
r = PchipInterpolator(knots, [0, 32, 237, 255, 0, 0])
g = PchipInterpolator(knots, [7, 107, 255, 170, 2, 7])
b = PchipInterpolator(knots, [100, 203, 255, 0, 0, 100])


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        '-q',
        '--quantization',
        type=int,
        default=2048,
        help='quantizatin rate')
    parser.add_argument(
        '-o',
        '--output',
        type=str,
        default='palette.go',
        help='output filename')
    args = parser.parse_args()
    return args


def build_palette(quantization):
    return [(int(r(t)), int(g(t)), int(b(t)))
            for t in np.linspace(0, 1, quantization)]


def dump(f, palette):
    f.write('package main\n\n')
    f.write('import "image/color"\n\n')
    f.write('var backgroundColor = color.RGBA{0x00, 0x00, 0x07, 0xff}\n\n')
    f.write('var palette = []color.Color{\n')

    for c in palette:
        f.write('\tcolor.RGBA{0x%02x, 0x%02x, 0x%02x, 0xff},\n' % c)

    f.write('}\n')


def main():
    args = parse_args()
    palette = build_palette(args.quantization)
    with open(args.output, 'w') as f:
        dump(f, palette)


if __name__ == '__main__':
    main()
