'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface Resource {
  id: string;
  url: string;
  normalized_url?: string;
  name?: string;
  category: string;
  categories?: string[];
  path?: string;
  source?: string;
  kind?: string;
  submodule?: boolean;
  researched: boolean;
  summary: string;
  features: string[];
  tags?: string[];
  docs_url?: string;
  homepage_url?: string;
  repo_url?: string;
  last_updated: string;
}

export default function ToolDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [resource, setResource] = useState<Resource | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) {
      setLoading(false);
      return;
    }

    const fetchResource = async () => {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch(`/api/resources/${id}`);
        if (!res.ok) {
          throw new Error(res.status === 404 ? 'Resource not found' : 'Failed to load resource');
        }
        const data = await res.json();
        setResource(data.resource ?? null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load resource');
      } finally {
        setLoading(false);
      }
    };

    fetchResource();
  }, [id]);

  if (loading) {
    return <div className="p-8 text-center">Loading resource details...</div>;
  }

  if (error) {
    return (
      <div className="p-8 space-y-4">
        <div className="text-lg font-semibold">{error}</div>
        <Button asChild variant="outline">
          <Link href="/dashboard/tools">Back to Tools</Link>
        </Button>
      </div>
    );
  }

  if (!resource) {
    return (
      <div className="p-8 space-y-4">
        <div className="text-lg font-semibold">Resource not found</div>
        <Button asChild variant="outline">
          <Link href="/dashboard/tools">Back to Tools</Link>
        </Button>
      </div>
    );
  }

  const categories = resource.categories?.length ? resource.categories : [resource.category];
  const links = [
    resource.homepage_url ? { label: 'Homepage', url: resource.homepage_url } : null,
    resource.repo_url ? { label: 'Repository', url: resource.repo_url } : null,
    resource.docs_url ? { label: 'Docs', url: resource.docs_url } : null,
    { label: 'Source URL', url: resource.url }
  ].filter(Boolean) as Array<{ label: string; url: string }>;

  return (
    <div className="p-6 space-y-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-3xl font-bold">{resource.name || resource.url}</h1>
          <p className="text-muted-foreground">{resource.normalized_url || resource.url}</p>
        </div>
        <Button asChild variant="outline">
          <Link href="/dashboard/tools">Back to Tools</Link>
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              {resource.summary || 'No research summary available yet.'}
            </p>
            <div className="flex flex-wrap gap-2">
              {categories.filter(Boolean).map(category => (
                <Badge key={category} variant="secondary">
                  {category}
                </Badge>
              ))}
              {resource.kind && <Badge variant="outline">{resource.kind}</Badge>}
              {resource.source && <Badge variant="outline">{resource.source}</Badge>}
            </div>
            {resource.features?.length > 0 && (
              <div className="space-y-2">
                <div className="text-sm font-semibold">Key Features</div>
                <div className="flex flex-wrap gap-2">
                  {resource.features.map(feature => (
                    <Badge key={feature} variant="secondary">
                      {feature}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Metadata</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Status</span>
              {resource.researched ? (
                <span className="text-green-500">Indexed</span>
              ) : (
                <span className="text-yellow-500">Pending</span>
              )}
            </div>
            {resource.path && (
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Submodule Path</span>
                <span className="text-right">{resource.path}</span>
              </div>
            )}
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Submodule</span>
              <span>{resource.submodule ? 'Yes' : 'No'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Last Updated</span>
              <span>{new Date(resource.last_updated).toLocaleString()}</span>
            </div>
            {resource.tags?.length ? (
              <div className="space-y-2">
                <div className="text-muted-foreground">Tags</div>
                <div className="flex flex-wrap gap-2">
                  {resource.tags.map(tag => (
                    <Badge key={tag} variant="outline">
                      {tag}
                    </Badge>
                  ))}
                </div>
              </div>
            ) : null}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Links</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {links.map(link => (
            <div key={link.url} className="flex flex-col gap-1">
              <span className="text-sm text-muted-foreground">{link.label}</span>
              <a
                href={link.url}
                target="_blank"
                rel="noreferrer"
                className="text-sm text-blue-400 hover:underline break-all"
              >
                {link.url}
              </a>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
