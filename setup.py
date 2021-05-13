from setuptools import find_packages
from setuptools import setup

import ufdn

setup(
    name='ufdn@hobbymarks',
    version='2021.05',
    packages=find_packages('.'),
    url='https://github.com/hobbymarks/UFn',
    license='MIT',
    author='HobbyMarks',
    author_email='ihobbymarks@gmail.com',
    description='',
    entry_points={
        'console_scripts': [
            'ufdn = ufdn:main',
        ],
    },
)
