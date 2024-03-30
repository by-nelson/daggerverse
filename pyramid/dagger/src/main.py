"""A module for testing and building complex Python applications."""

import dagger

from dagger import Doc, dag, field, function, object_type
from typing import Self, Annotated

@object_type
class Pyramid:
    """Python test and build related functions."""

    ctr: dagger.Container = field(default=(lambda: dag.container()))

    @function
    def with_version(
            self, 
            version: Annotated[str, Doc("Python container version to use")] = "3.11-bullseye",
    ) -> Self:
        """Initialize container with a given python version."""

        self.ctr = self.ctr.from_("python:" + version)
        return self

    @function
    def container(self) -> dagger.Container:
        """Get the container created by Pyramid."""
        return self.ctr

    @function
    def get_version(self) -> str:
        """Get the version created by Pyramid."""
        return (
            self.ctr.with_exec(["python3", "--version"])
            .stdout()
        )
