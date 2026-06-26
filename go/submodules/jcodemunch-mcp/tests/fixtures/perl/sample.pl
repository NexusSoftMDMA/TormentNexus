#!/usr/bin/perl
use strict;
use warnings;

package Animal;

use constant MAX_LEGS => 4;
use constant KINGDOM => 'Animalia';

# Create a new Animal
sub new {
    my ($class, %args) = @_;
    return bless { name => $args{name}, legs => $args{legs} // 4 }, $class;
}

# Get the animal's name
sub get_name {
    my ($self) = @_;
    return $self->{name};
}

=pod

=head1 describe

Returns a description of the animal.

=cut

sub describe {
    my ($self) = @_;
    return "Animal: " . $self->{name};
}

package Animal::Dog;

sub new {
    my ($class, %args) = @_;
    return bless { name => $args{name} }, $class;
}

sub bark {
    my ($self) = @_;
    return "Woof!";
}

1;
