"""A module for testing and building Python applications."""

import dagger

from dagger import Doc, dag, field, function, object_type
from typing import Self, Annotated, Optional

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
    def with_directory(
            self,
            directory: Annotated[dagger.Directory, Doc("Directory to run the formatter in")], 
    ) -> Self:
        """Sets the Pyramid working directory."""
        self.ctr = self.ctr.with_mounted_directory("/app", directory).with_workdir("/app")
        return self

    @function
    def with_pytest(self) -> Self:
        """Install pytest on the Pyramid container."""
        self.ctr = self.ctr.with_exec(["pip3", "install", "-U", "pytest"])
        return self

    @function
    def with_yapf(self) -> Self:
        """Install yapf on the Pyramid container."""
        self.ctr = self.ctr.with_exec(["pip3", "install", "-U", "yapf"])
        return self

    @function
    async def format(
            self,
            apply:   Annotated[bool, Doc("Whetever to run as test or format the code in place")] = False,
            verbose: Annotated[bool, Doc("Whetever to get verbose output")] = False,
    ) -> Self:
        """Use yapf to format or test code."""
        args = ["yapf", "--recursive", "--parallel"]

        if apply:
            args.append("--in-place")
        else:
            args.append("--diff")

        if verbose:
            args.append("--verbose")

        args.append(".")

        self.ctr = self.ctr.with_exec(args)
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

    @function
    async def get_directory(self) -> dagger.Directory:
        """Get the working directory."""
        directory = await self.ctr.workdir()
        return self.ctr.directory(directory)
