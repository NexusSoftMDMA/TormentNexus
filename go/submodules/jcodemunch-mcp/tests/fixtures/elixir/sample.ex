defmodule MyApp.Calculator do
  @moduledoc """
  A simple calculator module.
  """

  @type result :: {:ok, number()} | {:error, String.t()}

  @doc """
  Adds two numbers together.
  """
  @spec add(number(), number()) :: number()
  def add(a, b) do
    a + b
  end

  @doc """
  Subtracts b from a.
  """
  def subtract(a, b), do: a - b

  @doc false
  defp validate(x) when is_number(x) do
    {:ok, x}
  end

  defmacro debug(expr) do
    quote do
      IO.inspect(unquote(expr))
    end
  end
end

defmodule MyApp.Types do
  @moduledoc "Type definitions."

  @type name :: String.t()
  @typep age :: non_neg_integer()
  @opaque token :: binary()

  defguard is_positive(x) when is_number(x) and x > 0
end

defprotocol MyApp.Printable do
  @moduledoc "Protocol for printable types."

  @callback render(term()) :: String.t()

  @doc "Renders the value as a string."
  def to_string(value)
end

defimpl MyApp.Printable, for: Integer do
  def to_string(value) do
    Integer.to_string(value)
  end
end
