# Serialization helpers shared across models.
module Serializable
  def serialize
    instance_variables.each_with_object({}) do |var, hash|
      hash[var.to_s.delete('@')] = instance_variable_get(var)
    end
  end
end

# Represents a user in the system.
class User
  include Serializable

  ROLES = [:admin, :moderator, :user].freeze

  attr_accessor :name, :email

  # Initializes a new User.
  def initialize(name, email)
    @name = name
    @email = email
    @role = :user
  end

  # Finds a user by ID.
  def self.find(id)
    nil
  end

  # Returns a greeting string.
  def greet
    "Hello, #{@name}!"
  end

  private

  # Validates email format.
  def valid_email?
    @email.include?('@')
  end
end

# Standalone helper function.
def format_name(first, last)
  "#{first} #{last}"
end
