import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Highlight } from "./Highlight.tsx";

describe("Highlight", () => {
  it("renders plain text when query is empty", () => {
    render(<Highlight text="Hello world" query="" />);
    expect(screen.getByText("Hello world")).toBeInTheDocument();
    expect(screen.queryByRole("mark")).not.toBeInTheDocument();
  });

  it("renders plain text when query does not match", () => {
    render(<Highlight text="Hello world" query="xyz" />);
    expect(screen.getByText("Hello world")).toBeInTheDocument();
    expect(screen.queryByRole("mark")).not.toBeInTheDocument();
  });

  it("highlights matching substring", () => {
    render(<Highlight text="Hello world" query="world" />);
    expect(screen.getByRole("mark")).toHaveTextContent("world");
  });

  it("is case insensitive", () => {
    render(<Highlight text="Hello World" query="hello" />);
    expect(screen.getByRole("mark")).toHaveTextContent("Hello");
  });

  it("preserves original casing in highlighted text", () => {
    render(<Highlight text="FooBar" query="foobar" />);
    expect(screen.getByRole("mark")).toHaveTextContent("FooBar");
  });

  it("highlights only the first occurrence", () => {
    render(<Highlight text="test a test" query="test" />);
    const marks = screen.getAllByRole("mark");
    expect(marks).toHaveLength(1);
    expect(marks[0]).toHaveTextContent("test");
  });

  it("renders text before and after the match", () => {
    const { container } = render(<Highlight text="abc def ghi" query="def" />);
    expect(container).toHaveTextContent("abc def ghi");
    expect(screen.getByRole("mark")).toHaveTextContent("def");
  });

  it("handles match at the start of text", () => {
    const { container } = render(<Highlight text="Hello world" query="Hello" />);
    expect(screen.getByRole("mark")).toHaveTextContent("Hello");
    expect(container).toHaveTextContent("Hello world");
  });

  it("handles match at the end of text", () => {
    const { container } = render(<Highlight text="Hello world" query="world" />);
    expect(screen.getByRole("mark")).toHaveTextContent("world");
    expect(container).toHaveTextContent("Hello world");
  });

  it("handles full text match", () => {
    render(<Highlight text="exact" query="exact" />);
    expect(screen.getByRole("mark")).toHaveTextContent("exact");
  });
});
