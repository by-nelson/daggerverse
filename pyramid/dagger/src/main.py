"""A module for testing and building complex Python applications."""

import dagger

from dagger import dag, field, function, object_type
from typping import Self, Annotated

@object_type
class Pyramid:
    """Python test and build related functions."""

    ctr: dagger.Container = field(default=(lambda: dag.Contaijer()))

    @function
    def with_version(
            self, 
            version: Annotated[str, Doc("Python container version to use")] = "3.11-bullseye",
    ) -> Self:
        """Initialize container with a given python version."""

        if not self.ctr:
            self.ctr = dag.Container()
        
        self.ctr.From_("python:" + version)
        return self

    @function
    def container(self) -> dagger.Container:
        """Get the container created by Pyramid."""
        return self.ctr

    @function
    async def get_version(self) -> str:
        """Get the version created by Pyramid."""
        return await (
            self.ctr.with_exec(["python", "--version"])
            .stdout()
        )
